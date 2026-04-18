#include "jreng_value_tree_json.h"
#include <jreng_core/jreng_core.h>

namespace jreng::json
{ /*____________________________________________________________________________*/

static juce::ValueTree primitiveToChild (const juce::var& value)
{
    juce::ValueTree child { ID::item };
    child.setProperty (ID::value, value, nullptr);
    return child;
}

static juce::var childToPrimitive (const juce::ValueTree& child)
{
    return child.getProperty (ID::value);
}

static juce::ValueTree objectToValueTree (const juce::var& source,
                                           const juce::Identifier& rootId);

static juce::ValueTree arrayToValueTree (const juce::var& source,
                                          const juce::Identifier& rootId)
{
    jassert (source.isArray() and "arrayToValueTree: source must be an array");

    juce::ValueTree tree { rootId };
    tree.setProperty (ID::jsonArray, true, nullptr);
    const auto* arr { source.getArray() };

    for (int i { 0 }; i < arr->size(); ++i)
    {
        const juce::var& element { arr->getReference (i) };

        if (element.isObject())
        {
            tree.addChild (objectToValueTree (element, ID::item), -1, nullptr);
        }
        else if (element.isArray())
        {
            tree.addChild (arrayToValueTree (element, ID::item), -1, nullptr);
        }
        else
        {
            tree.addChild (primitiveToChild (element), -1, nullptr);
        }
    }

    return tree;
}

static juce::ValueTree objectToValueTree (const juce::var& source,
                                           const juce::Identifier& rootId)
{
    jassert (source.isObject() and "objectToValueTree: source must be an object");

    juce::ValueTree tree { rootId };

    if (auto* obj = source.getDynamicObject())
    {
        for (const auto& prop : obj->getProperties())
        {
            const juce::Identifier& key { prop.name };
            const juce::var& val { prop.value };

            if (val.isObject())
            {
                tree.addChild (objectToValueTree (val, key), -1, nullptr);
            }
            else if (val.isArray())
            {
                tree.addChild (arrayToValueTree (val, key), -1, nullptr);
            }
            else
            {
                tree.setProperty (key, val, nullptr);
            }
        }
    }

    return tree;
}

juce::ValueTree ValueTree::fromJson (const juce::var& source,
                                     const juce::Identifier& rootId)
{
    juce::ValueTree result { rootId };

    if (source.isObject())
    {
        result = objectToValueTree (source, rootId);
    }
    else if (source.isArray())
    {
        result = arrayToValueTree (source, rootId);
    }
    else
    {
        result.setProperty (ID::value, source, nullptr);
    }

    return result;
}

static juce::var valueTreeToObject (const juce::ValueTree& tree);

static juce::var valueTreeToArray (const juce::ValueTree& tree)
{
    juce::Array<juce::var> arr;

    for (int i { 0 }; i < tree.getNumChildren(); ++i)
    {
        const juce::ValueTree child { tree.getChild (i) };

        if (child.hasProperty (ID::value))
        {
            arr.add (childToPrimitive (child));
        }
        else if (child.hasProperty (ID::jsonArray))
        {
            arr.add (valueTreeToArray (child));
        }
        else
        {
            arr.add (valueTreeToObject (child));
        }
    }

    return juce::var { arr };
}

static juce::var valueTreeToObject (const juce::ValueTree& tree)
{
    juce::DynamicObject::Ptr obj { new juce::DynamicObject() };

    for (int i { 0 }; i < tree.getNumProperties(); ++i)
    {
        const juce::Identifier key { tree.getPropertyName (i) };
        obj->setProperty (key, tree.getProperty (key));
    }

    for (int i { 0 }; i < tree.getNumChildren(); ++i)
    {
        const juce::ValueTree child { tree.getChild (i) };

        if (child.hasProperty (ID::jsonArray))
        {
            obj->setProperty (child.getType(), valueTreeToArray (child));
        }
        else
        {
            obj->setProperty (child.getType(), valueTreeToObject (child));
        }
    }

    return juce::var { obj };
}

juce::var ValueTree::toJson (const juce::ValueTree& tree)
{
    jassert (tree.isValid() and "toJson: tree must be valid");

    juce::var result;

    if (tree.hasProperty (ID::jsonArray))
    {
        result = valueTreeToArray (tree);
    }
    else if (tree.hasProperty (ID::value))
    {
        result = tree.getProperty (ID::value);
    }
    else
    {
        result = valueTreeToObject (tree);
    }

    return result;
}

/**_____________________________END OF NAMESPACE______________________________*/
}// namespace jreng::json
